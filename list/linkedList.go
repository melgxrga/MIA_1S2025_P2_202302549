package lista

import (
	"fmt"
	"strconv"
	"strings"
	"bytes"
	"github.com/melgxrga/proyecto1Archivos/consola"
	"github.com/melgxrga/proyecto1Archivos/structures"
	"github.com/melgxrga/proyecto1Archivos/functions"
)

type MountNode struct {
	Key, Ruta   string
	Digits, Pos int
	Value       *datos.Partition
	ValueL      *datos.EBR
	Next, Prev  *MountNode
}


func (m *MountNode) MountNode(ruta string, digits int, pos int, value *datos.Partition, valueL *datos.EBR) {
	m.Ruta = ruta
	m.Digits = digits
	m.Pos = pos
	m.Value = value
	m.ValueL = valueL
	m.Next = nil
	m.Prev = nil
	m.CreateKey()
}

func (m *MountNode) CreateKey() string {
	directory := strings.Split(m.Ruta, "/")
	lastPart := directory[len(directory)-1]
	fileNameParts := strings.Split(lastPart, ".")
	filename := fileNameParts[0]
	return filename + strconv.Itoa(m.Pos) + strconv.Itoa(m.Digits)
}

type MountList struct {
	First, Last *MountNode
	Tamano      int
}

func (m *MountList) IsEmpty() bool {
	return m.First == nil
}

func (m *MountList) Mount(path string, digit int, part *datos.Partition, partL *datos.EBR) {
	newNode := &MountNode{
		Ruta:   path,
		Key:    "",
		Digits: digit,
		Pos:    m.CountPartitions(path),
		Value:  part,
		ValueL: partL,
	}
	newNode.Key = newNode.CreateKey()
	// fmt.Println(m.IsEmpty())
	if !m.IsEmpty() {
		m.Last.Next = newNode
		newNode.Prev = m.Last
		m.Last = newNode
		m.Tamano++
	} else {
		m.First = newNode
		m.First.Next = nil
		m.Last = newNode
		m.Last.Next = nil
		m.Tamano++
	}
	// deberia crear un m.PrintId o guardar en un singleton la consola
	m.GetId(newNode)
}
// GetNodeByPath busca el nodo de montaje cuyo path es prefijo del path solicitado (y de mayor longitud)
func (m *MountList) GetNodeByPath(absPath string) *MountNode {
	var best *MountNode
	maxLen := -1
	temp := m.First
	for temp != nil {
		if strings.HasPrefix(absPath, temp.Ruta) && len(temp.Ruta) > maxLen {
			best = temp
			maxLen = len(temp.Ruta)
		}
		temp = temp.Next
	}
	return best
}

// GetAllMountNodes devuelve todos los nodos de la lista de montajes
func (m *MountList) GetAllMountNodes() []*MountNode {
	var nodes []*MountNode
	temp := m.First
	for temp != nil {
		nodes = append(nodes, temp)
		temp = temp.Next
	}
	return nodes
}
func (m *MountList) UnMount(key_ string) *MountNode {
	if !m.IsEmpty() {
		temp := m.First
		counter := 0
		for counter < m.GetSize() {
			if key_ == temp.Key {
				if temp == m.First {
					m.First = m.First.Next
				} else if temp == m.Last {
					m.Last = m.Last.Prev
					m.Last.Next = nil
				} else {
					temp.Prev.Next = temp.Next
					temp.Next.Prev = temp.Prev
				}
				m.Tamano--
				return temp
			}
			temp = temp.Next      
			counter++            
		}
	}
	return nil
}

func (m *MountList) GetNodeById(key_ string) *MountNode {
	if m.IsEmpty() {
		fmt.Println("Depuración: La lista de montajes está vacía")
		return nil
	}

	fmt.Printf("Depuración: Buscando nodo con ID '%s'\n", key_)

	temp := m.First
	for temp != nil {
		fmt.Printf("Depuración: Comparando con nodo -> ID: '%s', Ruta: '%s'\n", temp.Key, temp.Ruta)
		if key_ == temp.Key {
			fmt.Println("Depuración: Nodo encontrado")
			return temp
		}
		temp = temp.Next
	}

	fmt.Println("Depuración: No se encontró el nodo")
	return nil
}


func (m *MountList) NodeExist(key_ string) bool {
	m.PrintList()
	if !m.IsEmpty() {
		temp := m.First
		for temp != nil {
			if key_ == temp.Key {
				return true
			}
			temp = temp.Next
		}
	}
	return false
}

func (m *MountList) CountPartitions(path string) int {
	contador := 1
	if !m.IsEmpty() {
		var temp *MountNode
		temp = m.First
		for temp != nil {
			if path == temp.Ruta {
				contador++
			}
			temp = temp.Next
		}
	}
	return contador
}

func (m *MountList) GetSize() int {
	return m.Tamano
}

func (m *MountList) GetId(node *MountNode) {
	consola.AddToConsole(fmt.Sprintf("Id: %s\n", node.Key))
}

func (m *MountList) PrintList() {
	str := ""
	for i := 0; i < 110; i++ {
		str += "-"
	}
	contenido := ""
	contenido += fmt.Sprintf("%s\n", str)
	contenido += fmt.Sprintf("%-15s", "Id")
	contenido += fmt.Sprintf("%-15s", "Name")
	contenido += fmt.Sprintf("%-10s", "Type")
	contenido += fmt.Sprintf("%-10s", "Fit")
	contenido += fmt.Sprintf("%-10s", "Start")
	contenido += fmt.Sprintf("%-10s", "Size")
	contenido += fmt.Sprintf("%-10s", "Status")
	contenido += fmt.Sprintf("%-30s\n", "Ruta")
	if !m.IsEmpty() {
		temp := m.First
		for temp != nil {
			contenido += fmt.Sprintf("%s\n", str)
			contenido += fmt.Sprintf("%-15s", temp.Key)
			if temp.Value != nil {
				contenido += fmt.Sprintf("%-15s", string(functions.TrimArray(temp.Value.Part_name[:])))
				contenido += fmt.Sprintf("%-10s", string(temp.Value.Part_type))
				contenido += fmt.Sprintf("%-10c", temp.Value.Part_fit)
				contenido += fmt.Sprintf("%-10d", temp.Value.Part_start)
				contenido += fmt.Sprintf("%-10d", temp.Value.Part_size)
				contenido += fmt.Sprintf("%-10c", temp.Value.Part_status)

			} else if temp.ValueL != nil {
				contenido += fmt.Sprintf("%-15s", string(functions.TrimArray(temp.ValueL.Part_name[:])))
				contenido += fmt.Sprintf("%-10s", "L")
				contenido += fmt.Sprintf("%-10c", temp.ValueL.Part_fit)
				contenido += fmt.Sprintf("%-10d", temp.ValueL.Part_start)
				contenido += fmt.Sprintf("%-10d", temp.ValueL.Part_size)
				contenido += fmt.Sprintf("%-10c", temp.ValueL.Part_status)
			}
			contenido += fmt.Sprintf("%-30s\n", temp.Ruta)
			temp = temp.Next
		}
	}
	contenido += fmt.Sprintf("%s\n\n", str)
	consola.AddToConsole(contenido)
}

var ListaMount = MountList{
	First:  nil,
	Last:   nil,
	Tamano: 0,
}

// LinkedList de usuarios
type UserID struct {
	uid   string
	gid   string
	uname string
	Next  *UserID
	Prev  *UserID
}


func (u *UserID) GetUID() string {
	return u.uid
}

func (u *UserID) GetGID() string {
	return u.gid
}

func (u *UserID) GetUName() string {
	return u.uname
}

type UserList struct {
	First  *UserID
	Last   *UserID
	Length int
}

func (u *UserList) IsEmpty() bool {
	return u.First == nil
}

func (u *UserList) AddUser(userId, groupId, username string) {
	newNode := &UserID{
		uid:   userId,
		gid:   groupId,
		uname: username,
	}
	if !u.IsEmpty() {
		u.Last.Next = newNode
		newNode.Prev = u.Last
		u.Last = newNode
		u.Length++
	} else {
		u.First = newNode
		u.First.Next = nil
		u.Last = newNode
		u.Last.Next = nil
		u.Length++
	}
}

func (u *UserList) GetUserById(userId_ string) *UserID {
	temp := u.First
	for temp != nil {
		if temp.uid == userId_ {
			return temp
		}
		temp = temp.Next
	}
	return nil
}

func (u *UserList) GetUsersByGroup(groupId_ string) []*UserID {
	var result []*UserID
	temp := u.First
	for temp != nil {
		if temp.gid == groupId_ {
			result = append(result, temp)
		}
	}
	return result
}

// IsMounted verifica si una partición ya está montada
func (m *MountList) IsMounted(path string, name [16]byte) bool {
    temp := m.First
    for temp != nil {
        // Verificar particiones primarias/extendidas
        if temp.Value != nil && bytes.Equal(temp.Value.Part_name[:], name[:]) {
            return true
        }
        // Verificar particiones lógicas
        if temp.ValueL != nil && bytes.Equal(temp.ValueL.Part_name[:], name[:]) {
            return true
        }
        temp = temp.Next
    }
    return false
}